""" Var að hakka smá í Pascal og þar er forrit sem reiknar hrópmerkt (factorial)
Hjakkaði þessu saman í Python og er alveg hræðilegur kóði (en virkar). Hægt að 
gera mun einfaldara. """

# Skrifað 1.5.2020
# Höf: Helgi Páll Jónsson

import numpy as np

def hropmerkt(tala):
    factorial = 1
    geymsla = []
    for i in range(1, tala + 1):
        geymsla.append([factorial*i])
        svar = np.prod(geymsla) # np.prod margfaldar saman gildinn í listanum
        return(svar)

number = int(input('Sláðu inn tölu og ég mun reikna hrópmerkt : '))
ans = print(hropmerkt(number))
print(ans)
