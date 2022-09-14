/* Write a program that will ask for a number of days 
and then will show how many hours, minutes and seconds 
are in that number of days */

// 31.5.2022
// Helgi Páll Jónsson

#include<stdio.h>

int main(void) {

    int days;
    float hours, minutes, seconds;

    printf("Number of days : ");
    scanf("%d", &days);

    hours = days * 24;
    minutes = hours * 60;
    seconds = minutes * 60;

    printf("\n");
    printf("  Hours: %.1f\n", hours);
    printf("Minutes: %.1f\n", minutes);
    printf("Seconds: %.1f\n\n", seconds);

    return 0;

}
