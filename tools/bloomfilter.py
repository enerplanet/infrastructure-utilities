import math
import hashlib
import urllib.request
import os
from typing import List

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

    def process(self, source: str, txt_output: str, bin_output: str):
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

        n = len(filtered_passwords)
        print(f"Found {len(raw_lines):,} raw entries -> Filtered down to {n:,} unique passwords (>= {self.min_length} chars).")

        if n == 0:
            print("No passwords matched the criteria. Aborting export.")
            return

        # 2. Export the filtered text list
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

        print("Success! Process completed.")


# --- How to Run It ---
if __name__ == "__main__":
    # Feel free to change this to a local path (e.g., "my_passwords.txt")
    DEFAULT_SOURCE = "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/100k-most-used-passwords-NCSC.txt"

    generator = UnifiedBloomGenerator(
        min_length=8,      # Your requested minimum length filter
        fp_rate=0.01       # Target false-positive rate (1%)
    )

    generator.process(
        source=DEFAULT_SOURCE,
        txt_output="ncsc_8plus_focused.txt",
        bin_output="../common/pkg/utils/password_filter.bin"
    )
