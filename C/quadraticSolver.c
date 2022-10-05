/* Forritlingur sem leysir 2. stigs margliður */
/* Höf: Helgi Páll Jónsson                    */
/* Dags: 5. október 2022                      */

#include <stdio.h>
#include <math.h>

float evaluateD(float A, float B, float C);

int main(void) {

  float A, B, C, D, posSol, negSol;

  // Hér er líka hægt að breyta taka innslátt frá notanda
  A = -1;
  B = 1;
  C = 1;

  D = evaluateD(A, B, C);

  if (D > 0)
  {
    posSol = ( -B + sqrt(D) ) / (2 * A);
    negSol = ( -B - sqrt(D) ) / (2 * A);
    printf("Lausn 1 = %lf\n", posSol);
    printf("Lausn 2 = %lf\n", negSol);
  }
  else if (D == 0)
  {
    posSol = ( -B + sqrt(D) ) / (2 * A);
    printf("Lausn = %lf\n", posSol);
  }
  else
  {
    printf("Engin lausn til!\n");
  }
  
  printf("B^2 - 4*A*C = %lf\n", D);
  
  return 0;
}

float evaluateD(float A, float B, float C)
{
  float D;
  D = B * B - 4 * A * C;
  return D;
}
