/* Ask for the users first name and then ask for their */
/* and display the output message Hello [First Name] [Surname] */

#include<stdio.h>
int main(void){

    char firstName[30];
    char surName[30];

    printf("User, what is your first name : ");
    scanf("%s", firstName);

    printf("User, what is your surname : ");
    scanf("%s", surName);

    printf("Hello %s %s", firstName, surName);

    return 0;

}
