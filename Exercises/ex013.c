/* Ask the user to enter a number that is under
20. If they enter a number that is 20 or 
more, display the message "Too high", 
otherwise display "Thank you". */

#include<stdio.h>

int main(void) {
  
  int number;
  
  print("Enter a number under 20 : ");
  scanf("%d", number);
  
  if (number > 20) {
    printf("Too high\n");
  } else printf("Thank you\n");
  
  return 0;
  
}
