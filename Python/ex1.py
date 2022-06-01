import matplotlib.pyplot as plt
import numpy as np


g = 9.8

x = []
y = []

for l in range(1,5000):
  T = 2 * np.pi * np.sqrt(l/g)
  y.append(T)
  x.append(l)

plt.plot(x,y)
plt.show()
