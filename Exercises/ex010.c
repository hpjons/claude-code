/* There are 2,204 pounds in a kilogram. Ask the */
/* user to enter weight in kilograms and convert it */
/* to pounds */

#include<stdio.h>

int main(void){

    float weight;

    printf("Enter weight in kilograms: ");
    scanf("%f", &weight);

    printf("\n\n%.1f kilograms = %.1f pounds\n\n", weight, weight * 2204);

    return 0;

}
