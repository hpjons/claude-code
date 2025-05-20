// fim. 16. jan. 2025
// Helgi Páll Jónsson
// Æfing í föllum og switch skipun
// Forritunamál: C

#include<stdio.h>

float reikna(int z, float x, float y);

int main()
{
  int svar = 0;
  float tala1 = 0.0;
  float tala2 = 0.0;

  printf("1. Samlagning\n");
  printf("2. Frádráttur\n");
  printf("3. Margföldun\n");
  printf("4. Deiling\n\n");
  
  printf("Veldu: ");
  scanf("%d", &svar);
  
  printf("Sláðu inn tölu 1: ");
  scanf("%f", &tala1);
  
  printf("Sláðu inn tölu 2: ");
  scanf("%f", &tala2);

  switch(svar) {
    case 1:    
      printf("Margfeldið er %.2f\n", reikna(svar, tala1, tala2));
    break;
    case 2:
      printf("Mismunurinn er %.2f\n", reikna(svar, tala1, tala2));
    break;
    case 3:
      printf("Margfeldið er %.2f\n", reikna(svar, tala1, tala2));
    break;
    case 4:
      printf("Kvótinn er %.2f\n", reikna(svar, tala1, tala2));
    break;
    default:
      printf("Eitthvað fór úrskeiðis!");
  }
}

float reikna(int z, float x, float y){
  
  float result = 0.0;
  
  switch (z){
    case 1:
      result = x + y;
      break;
    case 2: 
      result = x - y;
      break;
    case 3: 
      result = x * y;
      break;
    case 4:
      if (y != 0){
        result = x / y;
      } else 
        {
          printf("Viðvörun! Deiling með núlli er ekki leyfð.\n");
        }
      break;
    default:
      printf("Ógildur valkostur!\n");
  }
  return result;
}
