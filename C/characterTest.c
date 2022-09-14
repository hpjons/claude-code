#include <stdio.h>
#include <stdlib.h>
#include <math.h>

int main(void)
{
   int counter = 1000;

   FILE *fptr;

   fptr = fopen("WriteToMe.txt", "w");

   while (counter > 0)
   {
       fprintf(fptr, "nr = %d sqrt = %.4f cos = %.4f\n sin = %f.4",counter, sqrt(counter), cos(counter), sin ) ;
       counter = counter - 1;
   }

   fclose(fptr);

   return 0;
}
