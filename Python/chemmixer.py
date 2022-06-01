# Skrifað 13.5.2022
# Helgi Páll Jónsson
# Forritunarmál : Python

"""
Forrit sem segir þér hvernig þú blandar saman
tveimur efnum út frá gefnum hlutföllum.
"""

print('BLANDARINN - VELKOMINN')
print('Sláðu inn upplýsingar að neðan')
print('==============================')

chem1 = float(input('Hlutfall efnis 1 :'))
chem2 = float(input('Hlutfall efnis 2 : '))
volume = float(input('Heildarmagn efnablöndu : '))

total = chem1 + chem2

prop_chem1 = (volume/total) * chem1
prop_chem2 = (volume/total) * chem2

print('\nÞú blandar saman í eftirfarandi hlutföllum:')
print('Efni 1 samtals : ', round(prop_chem1, 2))
print('Efni 2 samtals : ', round(prop_chem2, 2))
print('\n')
