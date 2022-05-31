/* Ask the user to enter two numbers. Add them together and display the answer as 
The answer is [answer] */

// 30.5.2022
// Helgi Páll Jónsson

#include<stdio.h>

int main(void){

    int total, x, y;

    printf("User, enter two numbers please : \n");
    scanf("%d %d", &x, &y);

    total = x + y;

    printf("The total is %d", total);

    return 0;

}
