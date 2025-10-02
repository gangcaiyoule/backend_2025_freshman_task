import gradio as gr
import requests

session = requests.Session()
model_labels = ["deepseek-chat"]
print(session.cookies)
# 注册函数
def register(name, email, password):
    url = "http://localhost:8080/register"
    data = {
        "name": name,
        "email": email,
        "password": password
    }
    response = session.post(url, json=data)
    if response.status_code == 200:
        return f"注册成功！用户ID: {response.json().get('user_id')}"
    else:
        return f"注册失败: {response.json().get('error')}"

# 登录函数
def login(email, password):
    url = "http://localhost:8080/login"
    data = {
        "email": email,
        "password": password
    }
    response = session.post(url, json=data)
    if response.status_code == 200:
        return f"登录成功！"
    else:
        return f"登录失败: {response.json().get('error')}"

# 内容生成函数
def generate_content(prompt, history, model="deepseek-chat"):
    url = "http://localhost:8080/generate"
    data = {
        "prompt": prompt,
        "model": model
    }
    response = session.post(url, json=data)
    if response.status_code == 200:
        ai_response = response.json().get('result')
        history.append((prompt, ai_response))
        return history
    else:
        error = response.json().get('error', "生成失败")
        history.append((prompt, "❌ " + error))
        return history

def recharge():
    url = "http://localhost:8080/recharge"
    response = session.post(url)
    if response.status_code == 200:
        return "充值成功"
    else:
        return "充值失败"


# Gradio 界面
with gr.Blocks() as demo:
    gr.Markdown("# 用户注册、登录与内容生成")

    # 注册界面
    with gr.Row():
        name_input = gr.Textbox(label="用户名", placeholder="请输入用户名")
        email_input = gr.Textbox(label="邮箱", placeholder="请输入邮箱")
        password_input = gr.Textbox(label="密码", placeholder="请输入密码", type="password")
        register_button = gr.Button("注册")
        register_output = gr.Textbox(label="注册结果")
        register_button.click(register, inputs=[name_input, email_input, password_input], outputs=register_output)

    # 登录界面
    with gr.Row():
        login_email_input = gr.Textbox(label="邮箱", placeholder="请输入邮箱")
        login_password_input = gr.Textbox(label="密码", placeholder="请输入密码", type="password")
        login_button = gr.Button("登录")
        login_output = gr.Textbox(label="登录结果")
        login_button.click(
            login,
            inputs=[login_email_input, login_password_input],
            outputs=login_output
        )

    # 内容生成界面（改造成聊天框）
    chatbot = gr.Chatbot(label="对话历史")  # 聊天窗口
    prompt_input = gr.Textbox(label="生成提示", placeholder="请输入要生成的内容提示")
    generate_button = gr.Button("生成内容")

    # 点击按钮时：更新对话框
    generate_button.click(
        generate_content,
        inputs=[prompt_input, chatbot],
        outputs=chatbot
    )

    # 成为会员（充值）
    with gr.Row():
        recharge_button = gr.Button("成为会员")
        recharge_output = gr.Textbox(label="充值结果")
        recharge_button.click(recharge, outputs=recharge_output)

demo.launch(share=True)
