/*
Write a program that produces the following output:
For he's a jolly good fellow!
For he's a jolly good fellow!
For he's a jolly good fellow!
Which nobody can deny!

Have the program use two user-defined functions in addition to main(): one named
jolly() that prints the "jolly good" message once, and one named deny() that prints
the final line once.
*/

#include <stdio.h>

int jolly(void);
int deny(void);

int main(void)
{
    printf("For he's a");
    jolly();
    printf("fellow!\n");

    printf("For he's a");
    jolly();
    printf("fellow!\n");

    printf("For he's a");
    jolly();
    printf("fellow!\n");

    deny();

    return 0;
}

// The functions jolly() and deny()

int jolly(void)
{
    printf(" jolly good ");
    return 0;
}

int deny(void)
{
    printf("Which nobody can deny!");
    return 0;
}
