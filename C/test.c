#include <stdio.h>
float areaCircle(x);
int main(void)
{
    float ans;
    printf("Insert number : ");
    scanf("%f", ans);

    areaCircle(ans);

    return 0;
}

float areaCircle(x)
{
    float svar;
    svar = x * x * 3.14;
    printf("%f", svar);

    return;
}
