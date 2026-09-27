from openai import OpenAI
import os
import time

gpt_model = os.getenv("OPENAI_MODEL_NAME", "gpt-4o-mini")
openai_api_base_url = os.getenv("OPENAI_BASE_URL", "https://api.openai.com/v1")
openai_api_key = os.getenv("OPENAI_API_KEY")

total_input_tokens = 0


def SendSingleMsgToOpenAILLM(prompt: str):
    global total_input_tokens
    if not openai_api_key:
        raise RuntimeError("Set OPENAI_API_KEY before running the LLM judge.")

    client = OpenAI(
        api_key=openai_api_key,
        base_url=openai_api_base_url,
    )

    history = []
    history.append({"role": "user", "content": prompt})

    start_time = time.time()
    response = client.chat.completions.create(
        model=gpt_model,  # 指定模型
        messages=history,
        max_tokens=16384  # 最大生成的token数
    )
    end_time = time.time()
    interval = end_time - start_time
    print("时间间隔:", interval, "秒")
    if response == None:
        return None


    answer = response.choices[0].message.content
    usage = response.usage
    input_tokens = usage.prompt_tokens
    output_tokens = usage.total_tokens - input_tokens

    total_input_tokens += input_tokens
    print(f'[{gpt_model}] 本轮输入 tokens 量为：{input_tokens}')
    print(f'[{gpt_model}] 总输入 tokens 量为：{total_input_tokens}')
    

    return answer
