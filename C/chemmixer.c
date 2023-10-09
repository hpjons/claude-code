// maĂ­ 2022
// Helgi PĂĄll JĂłnsson

#include<stdio.h>
#include<math.h>
int main(void){
    
    float chem1, chem2, total, volume, prop_chem1, prop_chem2;

    printf("Hlutfall efnis 1 : ");
    scanf("%f", &chem1);
    printf("Hlutfall efnis 2 : ");
    scanf("%f", &chem2);
    printf("Heildarmagn efnablöndu : ");
    scanf("%f", &volume);

    total = chem1 + chem2;

    prop_chem1 = (volume/total) * chem1;
    prop_chem2 = (volume/total) * chem2;

    printf("\nÞú blandar saman í­ eftirfarandi hlutföllum:\n");
    printf("Efni 1 samtals : %f\n", prop_chem1);
    printf("Efni 2 samtals : %f\n", prop_chem2);

    return 0;

}
