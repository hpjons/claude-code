#include<stdio.h>

int main(void){

    int n;
    int failcount = 0;

    do
    {
        failcount = failcount + 1;
        printf("Sláðu inn heiltölu á milli 1 og 10: ");
        scanf("%d", &n);        
    }
    while (n != 3);

    printf("Til hamingju, þú giskaðir á rétta tölu eftir %d tilraunir\n", failcount);

    return 0;

}
