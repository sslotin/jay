import os
from random import randint

p = int(os.environ['SEED'])

print(randint(0, 10**p), randint(0, 10**p))
