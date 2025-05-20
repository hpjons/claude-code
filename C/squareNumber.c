#include<stdio.h>

int square(int toFunction);

int main()
{
	printf("Insert number : ");
	scanf("%i", number);
	square(number, x);
	printf("Square = %i", x);
}

int square(int toFunction, int fromFunction)
{
	fromFunction = toFunction * toFunction;
	return fromFunction;
}
