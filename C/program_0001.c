#include <stdio.h>

int main(void) {

    int time;
    int upperLimit;
    float speed;

    time = 0;
    speed = 0;
    upperLimit = 20;

    while (time < upperLimit) {
        speed = 9.8 * time;
        time++;
        printf("%.2f\n", speed);
    }

    return 0;
}
