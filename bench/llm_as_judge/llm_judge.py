import json
import os
import re
import random
from llm_client import SendSingleMsgToOpenAILLM
from datetime import datetime
from prompt import ABTEST_PROMPT_V2, SCORE_PROMPT_BPS, SCORE_PROMPT_COT, SCORE_PROMPT_DS, ABTEST_PROMPT

# the llm name
model_name = "llama"
# the llm answer file path
file_path = f"./llm_ans/{model_name}-judge.json"
# type should in ('chain', 'all', 'top1', 'top5', 'summary')
ans_type = "summary"
save_path = "./judge_res"
number = 2
a_type = "talk_chain_answer"
b_type = "all_answer"
# b_type = "assessor_vector_ans"


# llm-as-judge model, should in ('vote', 'score')
mode = "score"
ab_prompt_type = "cot"

judge_res = []

def post_process_string(input_string):
    # 1. 去掉前后空格
    input_string = input_string.strip()
    
    # 2. 去掉前后连续换行符
    input_string = input_string.strip('\n')
    
    return input_string


def startABJudge():
    print(f'当前开始 AB 测试，环境为 {a_type} vs {b_type}')
    global judge_res
    question_id = 0
    A = 0
    B = 0
    TIE = 0

    file = loadJsonObjFromFile(file_path)
    for case in file:
        
        case_id = case["case_id"]
        print(f"当前的 case_id 为 {case_id}")
        questions = case["questions"]
        for q in questions:
            question_id += 1
            question = q["question"]
            expected_answer = q["expected_answer"]
            a_answer = q[a_type]
            b_answer = q[b_type]

            full_prompt = ABTEST_PROMPT.format(question=question, expected_answer=expected_answer, a=a_answer, b=b_answer)
            if ab_prompt_type == "cot":
                full_prompt = ABTEST_PROMPT_V2.format(question=question, expected_answer=expected_answer, a=a_answer, b=b_answer)

            
            done = False
            time = 0
            answer = ''
            while not done and time < 5:
                time += 1

                raw_answer = SendSingleMsgToOpenAILLM(full_prompt)
                if answer == None:
                    continue
                answer = post_process_string(raw_answer)
                if ab_prompt_type == "cot":
                    answer = extract_judge_type(answer)
                    if answer not in ('[A]', '[B]', '[TIE]'):
                        print("COT格式不符合预期，重新生成")
                        continue
                # 3. 校验字符串是否只属于 '[A]' 或者 '[B]' 的其中一种
                if answer not in ('[A]', '[B]', '[TIE]'):
                    print("格式不符合预期，重新生成")
                    continue
                else:
                    if answer == '[A]':
                        A += 1
                    if answer == '[B]':
                        B += 1
                    if answer == '[TIE]':
                        TIE += 1
                    done = True
                    print(raw_answer)
            if not done:
                answer = '[输出有误]'
            print(f"当前问题是：{question}")
            print(f"{question_id} 当前选择的结果是：{answer}, 当前A/B/TIE比例为：{A}/{B}/{TIE}")
            
            judge_res.append({'id': question_id, 'quetsion': question, 'expected_answer': expected_answer, 'a_type': a_type, 'a_ans': a_answer, 'b_type': b_type, 'b_ans': b_answer, 'ai_judge': answer, 'raw_judge': raw_answer})

    print(f"ai-ab-judge done, model={model_name}, {a_type} vs {b_type}, 当前A/B/TIE比例为：{A}/{B}/{TIE}")
    judge_res.append({'A_votes': A, 'B_votes': B, 'Tie_votes': TIE})
    saveObjToJsonFile(save_path+'/ab/', model_name, judge_res, A/(A+B))
            

def startScoreJudge():
    global judge_res
    question_id = 0
    all_score = 0

    file = loadJsonObjFromFile(file_path)
    for case in file:
        print(f"cur plan: {model_name} - {ans_type}")
        case_id = case["case_id"]
        print(f"当前的 case_id 为 {case_id}")
        questions = case["questions"]
        for q in questions:
            question_id += 1
            question = q["question"]
            expected_answer = q["expected_answer"]
            talk_chain_answer = q["talk_chain_answer"]
            all_answer = q["all_answer"]
            top1_answer = q["top1_answer"]
            top5_answer = q["top5_answer"]
            summary_answer = q["summary_answer"]
            # chain_vector_answer = q["assessor_vector_ans"]


            full_prompt_chain = SCORE_PROMPT_BPS.format(question=question, expected_answer=expected_answer, candidate_answer=talk_chain_answer)
            full_prompt_all = SCORE_PROMPT_BPS.format(question=question, expected_answer=expected_answer, candidate_answer=all_answer)
            full_prompt_top1 = SCORE_PROMPT_BPS.format(question=question, expected_answer=expected_answer, candidate_answer=top1_answer)
            full_prompt_top5 = SCORE_PROMPT_BPS.format(question=question, expected_answer=expected_answer, candidate_answer=top5_answer)
            full_prompt_summary = SCORE_PROMPT_BPS.format(question=question, expected_answer=expected_answer, candidate_answer=summary_answer)

            ok_score = -1
            cur_prompt = ""
            cur_ans = ""

            if ans_type == "chain":
                cur_prompt = full_prompt_chain
                cur_ans = talk_chain_answer
            elif ans_type == "all":
                cur_prompt = full_prompt_all
                cur_ans = all_answer
            elif ans_type == "top1":
                cur_prompt = full_prompt_top1
                cur_ans = top1_answer
            elif ans_type == "top5":
                cur_prompt = full_prompt_top5
                cur_ans = top5_answer
            elif ans_type == "summary":
                cur_prompt = full_prompt_summary
                cur_ans = summary_answer

            done = False
            retry_times = 0

            while not done and retry_times < 5:
                retry_times += 1
                if ok_score == -1:
                    answer = SendSingleMsgToOpenAILLM(cur_prompt)
                    if answer is not None:
                        post_answer = post_process_string(answer)
                        score = extract_score(post_answer)
                        if score is not None:
                            ok_score = score
                            all_score += ok_score

                if ok_score != -1:
                    done = True
            print(answer)
            print(f"{question_id} - 当前评分为: {ok_score}")

            
            judge_res.append({'id': question_id, 'quetsion': question, 'expected_answer': expected_answer, 'candidate_ans_type': ans_type, 'ans_content': cur_ans, 'aijudge_score': ok_score, 'reasoning': answer})
    avg_score = all_score/question_id
    print(f"ai-score-judge done, avg score: {avg_score}")
    saveObjToJsonFile(save_path, model_name, judge_res, avg_score)

def extract_score(text):
    # 按行分割文本
    lines = text.strip().split('\n')
    # 找到 [Score] 标签的位置
    for i, line in enumerate(lines):
        if line.strip() == '[Score]':
            # 返回标签下的下一行内容并进行处理
            if i + 1 < len(lines):  # 确保下一行存在
                score_str = lines[i + 1].strip()
                # 校验提取的内容是否是整数
                if score_str.isdigit():
                    return int(score_str)
                else:
                    return None
    return None  # 如果没有找到 [Score] 标签，返回 None


def extract_judge_type(text):
    # 按行分割文本
    lines = text.strip().split('\n')
    # 找到 [Score] 标签的位置
    for i, line in enumerate(lines):
        if line.strip() == '[Judge]':
            # 返回标签下的下一行内容并进行处理
            if i + 1 < len(lines):  # 确保下一行存在
                judge_type = lines[i + 1].strip()
                return judge_type
    return None  # 如果没有找到 [Judge] 标签，返回 None

def saveObjToJsonFile(dir: str, model_name: str, obj, avg_score: float):
    middle_path = ""
    if mode == "vote":
        middle_path = mode + f'/{a_type}_vs_{b_type}'
    if mode == "score":
        middle_path = ans_type
    random_path = f'{dir}/{model_name}/{middle_path}/aijudge_{model_name}_{datetime.now().strftime("%m%d%H%M%S")}_{random.randint(10000, 99999)}_{round(avg_score, 2)}.json'
    print("结果保存到：", random_path)
    filePath = random_path
    os.makedirs(os.path.dirname(filePath), exist_ok=True)
    with open(filePath, 'w', encoding='utf-8') as f:
        json.dump(obj, f, indent=4, ensure_ascii=False)
        print("保存成功")


def loadJsonObjFromFile(filePath: str):
    with open(filePath, 'r', encoding='utf-8') as file:
        return json.load(file)
    

#startScoreJudge()
def main():
    i = 0
    # Run the configured evaluation once when invoked as a script.
    print(f"这是第 {i+1} 次循环")
    startABJudge()


if __name__ == "__main__":
    main()
