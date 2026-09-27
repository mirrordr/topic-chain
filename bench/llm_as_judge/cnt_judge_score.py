import argparse
import json

def loadJsonObjFromFile(filePath: str):
    with open(filePath, 'r', encoding='utf-8') as file:
        return json.load(file)
    

def main():
    parser = argparse.ArgumentParser(description="Calculate the average judge score.")
    parser.add_argument("path", nargs="?", default="./judge_res/example.json")
    args = parser.parse_args()

    res = loadJsonObjFromFile(args.path)
    scores = [item["aijudge_score"] for item in res]
    if not scores:
        raise ValueError("The result file contains no scores.")

    print("avg score: ", sum(scores) / len(scores))

if __name__ == "__main__":
    main()
