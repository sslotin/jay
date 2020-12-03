import sys

n = int(input())

l = 1
r = n
print("123", file=sys.stderr)

while l < r:
    m = (l + r) // 2
    print("?", m, flush=True)
    resp = input()
    if resp == ">":
        l = m + 1
    else:
        r = m

print("!", l, flush=True)
