#include<unistd.h>
#include<stdio.h>

void cough_n_sneeze(int n);

int main(void){

    cough_n_sneeze(10);

}

void cough_n_sneeze(int n)
{
    for(int i = 0; i < n; i++)
    {
        printf("%d Cough!!\n", i+1);
        sleep(1);
    }

    printf("Ahh..finally a break\n");
    sleep(3);
    printf("... oh no not again!!\n");


    for(int i = 0; i < n; i++)
    {
        printf("%d Atjsuu!!\n", i+1);
        sleep(1);
    }
}
