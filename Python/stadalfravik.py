import math

print('============================')
print('Þetta forrit tekur inn tölur')
print('og reiknar staðalfrávik')
print('============================\n')

fjoldi = int(input('Fjoldi : '))
numbers = []
for i in range(fjoldi):
    r = float(input('tala: '))
    numbers.append(r)

medaltal = sum(numbers)/fjoldi

print('Meðaltal = ', medaltal)

fravik = []

for item in numbers:
    fravik.append(item-medaltal)

print('Frávik frá meðaltali = ', fravik)

sq = []
for item in fravik:
    sq.append(item**2)

print('Frávik í öðru veldi = ', sq)

sum_sq = sum(sq)

print('Summa frávika í öðru veldi = ', sum_sq) 

calc = sum_sq/(fjoldi-1)
std = math.sqrt(calc)
std_rounded = round(std, 2)

print('Staðalfrávikið er = ', std_rounded)
