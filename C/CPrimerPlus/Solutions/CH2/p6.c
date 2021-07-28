/*
Write a program that creates an integer variable called toes. Have the program set toes
to 10. Also have the program calculate what twice toes is and what toes squared is. The
program should print all three values, identifying them.
*/

#include <stdio.h>
int main(void)
{
    int toes = 10;

    printf("toes = %d\ntoes x 2 = %d\ntoes squared = %d\n\n" ,toes, toes * 2, toes * toes);
    return 0;
}
