# -*- coding: utf-8 -*-
"""
Spyder Editor

This is a temporary script file.
"""

# Höf: Helgi Páll Jónsson
# Dags: 6.10.2022

rows = 500
cols = 6
a = 0
i = 0
counter = 0

while a < rows:
    for i in range(cols):
        print(i + counter, end='\t')
    print('\n')
    counter = counter + cols
    a = a + 1
