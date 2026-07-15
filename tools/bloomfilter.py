import math
import hashlib
import urllib.request
import os
import random
from typing import List, Optional
from itertools import permutations

class UnifiedBloomGenerator:
    def __init__(self, min_length: int = 8, fp_rate: float = 0.01):
        self.min_length = min_length
        self.fp_rate = fp_rate

    def _load_raw_passwords(self, source: str) -> List[str]:
        """Loads lines from either a local file path or a URL."""
        if source.startswith(("http://", "https://")):
            print(f"Downloading source from URL: {source} ...")
            try:
                with urllib.request.urlopen(source) as response:
                    content = response.read().decode('utf-8')
                    return content.splitlines()
            except Exception as e:
                raise RuntimeError(f"Failed to fetch URL {source}: {e}")
        else:
            print(f"Reading local source file: {source} ...")
            if not os.path.exists(source):
                raise FileNotFoundError(f"Local file not found at: {source}")
            with open(source, "r", encoding="utf-8", errors="ignore") as f:
                return f.readlines()

    def _get_hashes(self, item: str, m: int, k: int):
        """Generates k indices using Kirsch-Mitzenmacher double-hashing."""
        digest = hashlib.sha256(item.encode('utf-8')).digest()
        hash_a = int.from_bytes(digest[0:4], byteorder='big')
        hash_b = int.from_bytes(digest[4:8], byteorder='big')

        for i in range(k):
            yield (hash_a + i * hash_b) % m

    def _generate_compound_guesswords(self, guesswords: List[str]) -> List[str]:
        """Generate all pairwise combinations of guesswords (e.g. hello+world, world+hello)."""
        if len(guesswords) < 2:
            return []

        combined = set()
        for a, b in permutations(guesswords, 2):
            candidate = a + b
            if len(candidate) >= self.min_length:
                combined.add(candidate)

        result = sorted(combined)
        print(f"Generated {len(result):,} compound guesswords from {len(guesswords)} base words.")
        return result

    def process(self, source: str, export_txt: bool,txt_output: str, bin_output: str, custom_guesswords: Optional[List[str]] = None):
        # 1. Load and filter the raw passwords
        raw_lines = self._load_raw_passwords(source)

        # Filter duplicates and check minimum length
        seen = set()
        filtered_passwords = []
        for line in raw_lines:
            cleaned = line.strip()
            if len(cleaned) >= self.min_length and cleaned not in seen:
                seen.add(cleaned)
                filtered_passwords.append(cleaned)

        # 1b. Merge custom compound guesswords into the filtered list
        if custom_guesswords:
            compounds = self._generate_compound_guesswords(custom_guesswords)
            added = 0
            for pw in compounds:
                if pw not in seen:
                    seen.add(pw)
                    filtered_passwords.append(pw)
                    added += 1
            if added:
                print(f"Merged {added} additional compound guesswords into the filter list.")

        n = len(filtered_passwords)
        print(f"Found {len(raw_lines):,} raw entries -> Filtered down to {n:,} unique passwords (>= {self.min_length} chars).")

        if n == 0:
            print("No passwords matched the criteria. Aborting export.")
            return

        # 2. Export the filtered text list (optional, bin is always exported)
        if export_txt:
            print(f"Saving filtered text list to '{txt_output}'...")
            with open(txt_output, "w", encoding="utf-8") as f:
                for pw in filtered_passwords:
                    f.write(pw + "\n")

        # 3. Compute optimal Bloom Filter parameters dynamically
        # m = - (n * ln(p)) / (ln(2)^2)
        m = int(round(- (n * math.log(self.fp_rate)) / (math.log(2) ** 2)))
        # k = (m / n) * ln(2)
        k = int(round((m / n) * math.log(2)))

        num_bytes = (m + 7) // 8
        bit_array = bytearray(num_bytes)

        print(f"Configuring Bloom Filter (Error Rate: {self.fp_rate*100}%):")
        print(f" - Size: {m:,} bits ({num_bytes / 1024:.2f} KB)")
        print(f" - Hash Functions (k): {k}")

        # 4. Populate the Bit Array
        print("Populating Bloom filter bit array...")
        for pw in filtered_passwords:
            for bit_index in self._get_hashes(pw, m, k):
                byte_index = bit_index // 8
                bit_position = bit_index % 8
                bit_array[byte_index] |= (1 << bit_position)

        # 5. Export to binary Bloom Filter with metadata headers
        print(f"Saving compiled binary filter to '{bin_output}'...")
        with open(bin_output, 'wb') as f:
            # Metadata: 4 bytes for count (n), 8 bytes for bit-size (m), 1 byte for hashes (k)
            f.write(n.to_bytes(4, byteorder='big'))
            f.write(m.to_bytes(8, byteorder='big'))
            f.write(k.to_bytes(1, byteorder='big'))
            f.write(bit_array)

        # 6. Self-test: verify the filter catches a random sample
        self._self_test(filtered_passwords, m, k, bit_array)

        print("Success! Process completed.")

    def _self_test(self, passwords: List[str], m: int, k: int, bit_array: bytearray):
        """Pick 50 random passwords and verify they all pass the bloom filter."""
        sample_size = min(50, len(passwords))
        sample = random.sample(passwords, sample_size)

        misses = 0
        for pw in sample:
            for bit_index in self._get_hashes(pw, m, k):
                byte_idx = bit_index // 8
                bit_pos = bit_index % 8
                if not (bit_array[byte_idx] & (1 << bit_pos)):
                    misses += 1
                    break

        if misses == 0:
            print(f"Self-test passed: all {sample_size} sampled passwords found in the filter.")
        else:
            print(f"Self-test WARNING: {misses}/{sample_size} sampled passwords were NOT found in the filter.")


# --- How to Run It ---
if __name__ == "__main__":
    # Feel free to change this to a local path (e.g., "my_passwords.txt")
    DEFAULT_SOURCE = "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/100k-most-used-passwords-NCSC.txt"

    # Custom guesswords: all pairwise combinations are generated and merged
    # into the filter. Add words that users commonly combine in passwords.
    CUSTOM_GUESSWORDS = [
        "admin", "password", "welcome",
        "123", "abc", "test", "guest", "user", "enerplanet", "ener", "planet",
        "pass", "key", "login", "secure", "demo",
    ]

    generator = UnifiedBloomGenerator(
        min_length=10,      # Your requested minimum length filter
        fp_rate=0.01       # Target false-positive rate (1%)
    )

    generator.process(
        source=DEFAULT_SOURCE,
        export_txt=False,
        txt_output="ncsc_8plus_focused.txt",
        bin_output="../common/pkg/utils/password_filter.bin",
        custom_guesswords=CUSTOM_GUESSWORDS,
    )
