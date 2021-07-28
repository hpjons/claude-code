#include<stdio.h>
int main(){

    FILE * fpointer = fopen("kina.txt", "w");

    int num;

    for (num = 1; num < 15000; num++){
            fprintf(fpointer, "[%d] \t 20000 20000 20000 20000 20000\n", num);

    }


    fclose(fpointer);
    return 0;

}
