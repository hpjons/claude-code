#include<stdio.h>
#include<stdlib.h>
#include<math.h>

int main(void)
{
	int counter;
	
	FILE *intofile;
	
	intofile = fopen("cosines.txt", "a");
	
	for (counter = 0; counter < 100; counter++)
	{
		fprintf(intofile, "Cos(%d) = %f\n",counter, cos(counter));
	}
	
	return 0;
}
