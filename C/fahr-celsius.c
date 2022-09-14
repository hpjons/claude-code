#include<stdio.h>
int main()
{
	float fahr, celsius, kelvin;
	int lower, upper, step;

	lower = 0;
	upper = 300;
	step = 20;

	fahr = lower;


	printf("Fahr \t Cels \t\t Kelv\n");

	while (fahr <= upper) {
		celsius = (5.0/9.0) * (fahr-32.0);
		kelvin = celsius + 273;
		printf("%3.0f \t %6.1f \t %6.1f \n", fahr, celsius, kelvin);
		fahr = fahr + step;
	}
}
