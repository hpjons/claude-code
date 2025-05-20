spenna = lambda I, R : I * R
straumur = lambda U, R : U / R
vidnam = lambda I, U : U / I

print(f'Gefðu upp gildin sem þú veist en settu inn núll fyrir það gildi sem þú veist ekki:')
U = float(input('spenna (U) = '))
I = float(input('straumur (I) = '))
R = float(input('viðnám = (R) = '))

if U == 0:
    print(f'Spennan er {spenna(I, R)} volt eða {(spenna(I, R)/1000)} millivolt')
elif I == 0:
    print(f'Straumurinn er {straumur(U, R)} amper eða {straumur(U, R) * 1000} milliamper')
elif R == 0:
    print(f'Viðnámið er {vidnam(U, I)} ohm eða {vidnam(U, I)/1000} milliohm')





