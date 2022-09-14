/*
Write a program that produces the following output:
Brazil, Russia, India, China
India, China,
Brazil, Russia

Have the program use two user-defined functions in addition to main(): one named
br() that prints "Brazil, Russia" once, and one named ic() that prints "India, China"
once. Let main() take care of any additional printing tasks.
*/

#include <stdio.h>

int br(void);
int ic(void);

int main(void)
{
    br();
    ic();
    printf("India, China,\n");
    printf("Brazil, Russia");

    return 0;
}

int br(void)
{
    printf("Brazil, Russia ");
    return 0;
}

int ic(void)
{
    printf("India, China\n");
    return 0;
}
