# Teningakast
# 19.2.2023
# Helgi Páll Jónsson

import random

s1 = 0
s2 = 0
s3 = 0
s4 = 0
s5 = 0
s6 = 0

# Hér er sett inn hversu oft teningi skal kastað
numberOfThrows = 6

for i in range(numberOfThrows):
    coin_sides = [1,2,3,4,5,6]
    throw = random.choice(coin_sides)

    if throw == 1:
        s1 += 1
    elif throw == 2:
        s2 += 1
    elif throw == 3:
        s3 += 1
    elif throw == 4:
        s4 += 1
    elif throw == 5:
        s5 += 1
    elif throw == 6:
        s6 += 1

print('Fjöldi kasta var', numberOfThrows, 'og eftirfarandi hliðar komu upp:')
print('-----------------------------------------------------')
print('Hlið 1 : ', s1, 'sinnum eða', (s1/numberOfThrows)*100, '%')
print('Hlið 2 : ', s2, 'sinnum eða', (s2/numberOfThrows)*100, '%')
print('Hlið 3 : ', s3, 'sinnum eða', (s3/numberOfThrows)*100, '%')
print('Hlið 4 : ', s4, 'sinnum eða', (s4/numberOfThrows)*100, '%')
print('Hlið 5 : ', s5, 'sinnum eða', (s5/numberOfThrows)*100, '%')
print('Hlið 6 : ', s6, 'sinnum eða', (s6/numberOfThrows)*100, '%')
