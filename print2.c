/* print2.c-more printf() properties */
#include <stdio.h>
int main(void)
{
	unsigned int un = 3000000000; /* system with 32-bit int */
	short end = 200;			  /* and 16-bit short */
	long big = 65537;
	long long verybig = 12345678908642;
	char myName[7] = "Python";
	
	
	printf("un = %u and not %d\n", un, un);
	printf("end = %hd and %d\n", end, end);
	printf("big = %ld and not %hd\n", big, big);
	printf("verybig = %lld and not %ld\n", verybig, verybig);
	
	printf("%s", myName);
	
	return 0;
}
