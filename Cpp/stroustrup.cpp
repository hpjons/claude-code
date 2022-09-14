#include <stdio.h>
#include <string.h>

#define DEPT_TO_CHECK "HRD"

int main()
{
	char deptDetails[]="HRD: Mr. ABC";
	if(memcmp(deptDetails,DEPT_TO_CHECK,3)==0)
	{
		printf("Dept. name: Human Resource Department, Manager: %s\n",deptDetails+5);
	}
	else{
		printf("Details not found!!!\n");
	}

	return 0;
}
