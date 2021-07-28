#include<stdio.h>
#include<stdlib.h>
int main(){

    char grade = 'A';

    switch(grade){
    case 'A' :
        printf("Great");
        break;
    case 'B' :
        printf("Alright");
        break;
    case 'C' :
        printf("bad");
        break;
    case 'D' :
        printf("Total crap!");
        break;
    }

    return 0;
}
