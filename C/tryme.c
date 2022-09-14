#include <stdio.h>

int velocity(void);

int main(void)
{
    velocity();
    return 0;
}

int velocity(void)
{
    float v, v0, t, g;

    g = 9.8;

    printf("v0 (m/s) : ");
    scanf("%f", &v0);
    printf("t (s) : ");
    scanf("%f", &t);

    v = v0 + g * t;

    printf("%4.f (m/s)", v);

    return 0;

}
