/* Ask for two numbers. if
the first one is larger
than the second, display
the second number first
and then the first
number, otherwise show
the first number first and
then the second */

// 31.5.2022
// Helgi Páll Jónsson

#include<stdio.h>

int main(void){

    int first, second;

    printf("Give two numbers : ");
    scanf("%d %d", &first, &second);

    printf("\nIf first given > second given, second is displayed");
    printf("\nIf first given < second given, first is displayed\n\n");

    if (first > second) {
        printf("first --> %d\nsecond --> %d\n\n", second, first);
    } else printf("first --> %d\nsecond --> %d\n\n", first, second);

    return 0;

}
