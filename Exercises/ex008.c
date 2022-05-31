/* Ask for the total price of the bill. 
Divide the total bill by the number of 
diners and show how much each person must pay */

// 31.5.2022
// Helgi Páll Jónsson

#include<stdio.h>
int main(void){

    float price, diners;

    printf("Total price of bill : ");
    scanf("%f", &price);
    printf("Number of diners : ");
    scanf("%f", &diners);

    printf("Each diner pays %.1f", price/diners);

    return 0;
}
