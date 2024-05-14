import sys

def main():
    if len(sys.argv) != 3:
        print("Usage: python script.py <num1> <num2>")
        return

    try:
        num1 = int(sys.argv[1])
        num2 = int(sys.argv[2])
        result = num1 + num2
        print("result:", result)
    except ValueError:
        print("mast int")

if __name__ == "__main__":
    main()
