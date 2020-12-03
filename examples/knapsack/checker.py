import sys

def parse_ints(s: str):
    return list(map(int, s.split()))

with open(sys.argv[1], "r") as input_file:
    lines = input_file.readlines()
    n = int(lines[0])
    a = parse_ints(lines[1])

with open(sys.argv[2], "r") as output_file:
    k = parse_ints(output_file.readlines()[0])
    b = [a[i - 1] for i in k]

assert sum(b) % n == 0
assert len(k) > 0
assert len(set(k)) == len(k)
assert all(1 <= x <= n for x in k)
