#include<stdio.h>
#include<math.h>

int main()
{
    int counter;
    for (counter = 10; counter > 0; counter--) {
        printf("%.3f\n", pow(counter, 2));
    }
    return 0;
}

