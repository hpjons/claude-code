#include<stdio.h>
#include<math.h>

int main(){

int i;

for(i = 0; i <= 200000000; i++){
    printf("%2.f \t %2.f\n", pow(i, 2), pow(i, 3));
}

return 0;

}
