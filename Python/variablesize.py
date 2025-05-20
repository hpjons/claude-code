import sys

def print_type_limits(data_type):
  """
  Prints the upper and lower bounds of a given data type.

  Args:
    data_type: The data type to check (e.g., 'int', 'float').
  """

if data_type == 'int':
  print(f"Limits for {data_type}:")
  print(f"  Minimum: {sys.maxsize}")
  print(f"  Maximum: {-sys.maxsize - 1})
