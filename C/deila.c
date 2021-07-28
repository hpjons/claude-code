#include<stdio.h>
int main()
{
  int i;

  for (i = 1996; i <= 2021; i++) {
    if (i % 4 == 0)
    {
      printf("%i hlaupár \n", i);
    }
    else
      printf("%i\n", i);
  }
  return 0;
}
