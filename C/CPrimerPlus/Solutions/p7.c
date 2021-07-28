/*
Many studies suggest that smiling has benefits. Write a program that produces the
following output:

Smile!Smile!Smile!
Smile!Smile!
Smile!

Have the program define a function that displays the string Smile! once, and have the
program use the function as often as needed.
*/

#include <stdio.h>
int bros(void);

int main(void)
{
    bros();
    bros();
    bros();
    printf("\n");
    bros();
    bros();
    printf("\n");
    bros();
    return 0;
}

int bros(void)
{
    printf("Smile!");
    return 0;
}
