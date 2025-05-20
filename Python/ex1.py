celsius = -273
while celsius <= 30:
    kelvin = celsius + 273.15
    fahrenheit = (9.0/5.0) * celsius + 32

    print(f'{celsius:.1f} ˚C \t {kelvin:.1f} K \t {fahrenheit:.1f} ˚F')
    celsius += 1

