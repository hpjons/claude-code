#include <stdio.h>
int main()
{
    float cels, fahr, step;

    fahr = -25.0;
    step = 1.0;

    while (fahr <= 212)
    {
        cels = (fahr - 32.0) * (5.0/9.0);
        fahr = fahr + step;
        printf("%2.f °F --> %2.f °C\n", fahr, cels);
    }

    return 0;
}
