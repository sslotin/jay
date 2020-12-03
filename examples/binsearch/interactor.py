import os
import sys
import random

random.seed(os.environ['SEED'])

x = random.randint(1, 1000)
print(x, flush=True)

request_count = 0

while True:
    c, y = input().split()
    y = int(y)

    if c == '!':
        assert x == y, 'Wrong answer'
        break
    elif x > y:
        print('>', flush=True)
    elif x < y:
        print('<', flush=True)
    else:
        print("=", flush=True)

    request_count += 1
    assert request_count <= 20, "Too many requests made"
