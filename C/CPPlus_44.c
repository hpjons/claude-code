#include <stdio.h>
void butler(void);      /* ANSI/ISO C function prototyping */
int main(void)
{
    printf("I will summon the butler function.\n");
    butler();
    printf("Yes. Bring me some tea and writeble DVDs.\n");

    return 0;
}

void butler(void)      /* start of function definition */
{
    printf("You rang, sir?\n");
}
