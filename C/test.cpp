#include<stdio.h>
#include<math.h>

int main()
{
    float i;

    for(i = 0; i < 10; i++)
    {
        printf("%.2f ", cos(3*i));

    }

    printf("\n\n");

    for(i = 0; i < 10; i++)
    {
        printf("%.2f ", sin(2*i));
    }

   return 0;
}
