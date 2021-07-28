/*
In C, one function can call another. Write a program that calls a function named one_
three(). This function should display the word one on one line, call a second function
named two(), and then display the word three on one line. The function two() should
display the word two on one line. The main()
*/

#include <stdio.h>
int one_three(void);
int two(void);

int main(void)
{
    one_three();
    return 0;
}

int one_three()
{
    printf("one\n");
    two();
    printf("three");
    return 0;
}

int two(void)
{
    printf("two\n");
    return 0;
}
