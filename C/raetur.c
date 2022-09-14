#include <stdio.h>
#include <math.h>
int main()
{
	float counter;

	for (counter = 0; counter < 10; counter++)
	{
		printf("%f\n", sqrt(counter));
	}

	return 0;
}
