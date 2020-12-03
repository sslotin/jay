import os
import random

n = 2 + int(os.environ['SEED'])

print(n)

for _ in range(n):
    x = random.randint(0, n - 1)
    print(x, end=" ")
