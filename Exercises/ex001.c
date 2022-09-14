/* Ask for the user's first name and display the output message
Hello [First Name] */

// 30.5.2022
// Helgi Páll Jónsson

#include<stdio.h>
int main(){

    char name[30];
    printf("What is your name : ");
    scanf("%s", name);
    printf("Hello %s", name);

    return 0;

}
