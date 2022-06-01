#include<stdio.h>

int main(void) {
  
  int element;
  
  printf("Give element number : ");
  scanf("%d", &element);

  switch (element) {
    case 1: printf("Hydrogen"); break; case 2: printf("Helium"); break; case 3: printf("Lithium"); break;
    case 4: printf("Beryllium"); break; case 5: printf("Boron"); break; case 6: printf("Carbon"); break; 
    case 7: printf("Nitrogen"); break; case 8: printf("Oxygen"); break; case 9: printf("Fluorine"); break;
    case 8: printf("Neon"); break;
  }

return 0;

}
