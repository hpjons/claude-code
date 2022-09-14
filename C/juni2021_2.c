/* Bý til tvö föll sem taka breytu. Þau reikna rúmmál og flatarmál kúlu. Bý svo til lykkju sem reiknar fyrir
ákveðin fjölda af kúlum með mismunandi radíus
*/

#include<stdio.h>
#include<math.h>

double rummal(double tala)
{
    double ans = (4/3) * 3.14159265 * pow(tala, 3);
    return ans;
}

double flatarmal(double tala)
{
    double ans = 4 * 3.141569265 * pow(tala, 2);
    return ans;
}

int main()
{

    const int fasti = 5;

    int teljari;

    printf("VOLUME OF A SPHERE\n");
    printf("==================\n");
    for (teljari = 1; teljari <= fasti; teljari++)
    {
        printf("Volume of sphere with radius %lf is %lf\n", teljari/1.0, rummal(teljari) );
    }

    printf("\n\nSURFACE OF A SPHERE\n");
    printf("===================\n");
    for (teljari = 1; teljari <= fasti; teljari++)
    {
        printf("Surface of sphere with radius %lf is %lf\n", teljari/1.0, flatarmal(teljari) );

    }
return 0;
}
