/* Task the user to enter a number over 100 and then enter a number under
10 and tell them how many times the smaller number goes into the larger
number in a user-friendly format. */

// 31.5.2022
// Helgi Páll Jónsson

#include<stdio.h>
int main(void){

    float lower, upper;

    printf("Enter a number over 100 : ");
    scanf("%f", &upper);
    printf("Enter a number under 10 : ");
    scanf("%f", &lower);

    printf("\nThe number %.1f goes %.1f times into the number %.1f\n", lower, upper/lower, upper);

    return 0;

}
