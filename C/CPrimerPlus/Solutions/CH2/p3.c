/*
Write a program that converts your age in years to days and displays both values. At this
point, don't worry about fractional years and leap years.
*/
#include <stdio.h>
int main(void)
{
    int myAgeYears = 44;

    int myAgeDays = myAgeYears * 365;

    printf("Age in years : %d\n", myAgeYears);
    printf("Age in days  : %d\n", myAgeDays);

    return 0;
}
