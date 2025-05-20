#include <stdio.h>

int main(void)
{
    for (int i = 2024; i < 2045; i++)
    {
        if (i % 4 == 0)
        {
            printf("%i LEAP YEAR\n", i);
        }
        else
        {
            printf("%i\n", i);
        }
    }
    return 0;
}