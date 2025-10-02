# import gradio as gr
# import requests
#
#
# model_labels = ["deepseek-chat"]
# # 注册函数
# session_token = None
# def register(name, email, password):
#     url = "http://localhost:8080/register"
#     data = {
#         "name": name,
#         "email": email,
#         "password": password
#     }
#     response = requests.post(url, json=data)
#     if response.status_code == 200:
#         return f"注册成功！用户ID: {response.json().get('user_id')}"
#     else:
#         return f"注册失败: {response.json().get('error')}"
#
# # 登录函数
# def login(email, password):
#     global session_token
#     url = "http://localhost:8080/login"
#     data = {
#         "email": email,
#         "password": password
#     }
#     response = requests.post(url, json=data)
#     session_token = response.json().get('token')
#     if response.status_code == 200:
#         return f"登录成功！Token: {session_token}"
#
#     else:
#         return f"登录失败: {response.json().get('error')}"
#
# # 内容生成函数
# def generate_content(prompt, history, model, token):
#     url = "http://localhost:8080/generate"
#     headers = {
#         "Authorization": f"Bearer {token}"
#     }
#     data = {
#         "prompt": prompt,
#         "model": model
#     }
#     response = requests.post(url, json=data, headers=headers)#前端->后端
#     #在对话框展示先前聊天记录
#     if response.status_code == 200:
#         ai_response = response.json().get('result')#后端->前端
#         history.append((prompt, ai_response))
#         return history
#     else:
#         error = response.json().get('error', "生成失败")
#         history.append((prompt, "❌ " + error))
#         return history
#
# def recharge(token):
#     url = "http://localhost:8080/recharge"
#     headers = {
#         "Authorization": f"Bearer {token}"
#     }
#     response = requests.post(url, headers=headers)
#     if response.status_code == 200:
#         return "充值成功"
#     else:
#         return "充值失败"
#
# # #获取模型
# # def get_models(token):
# #     url = "http://localhost:8080/getModel"
# #     headers = {
# #         "Authorization": f"Bearer {token}"
# #     }
# #     response = requests.get(url, headers=headers)
# #     #查询可用的模型名
# #     if response.status_code == 200:
# #         labels = response.json().get('availableModel')
# #         return labels
# #     else:
# #         print(response.json().get('error', "查询可用模型名失败"))
# #         return ["deepseek-chat"]
#
# # model_labels = get_models(session_token)
# # Gradio 界面
# with gr.Blocks() as demo:
#     gr.Markdown("# 用户注册、登录与内容生成")
#
#     # 注册界面
#     with gr.Row():
#         name_input = gr.Textbox(label="用户名", placeholder="请输入用户名")
#         email_input = gr.Textbox(label="邮箱", placeholder="请输入邮箱")
#         password_input = gr.Textbox(label="密码", placeholder="请输入密码", type="password")
#         register_button = gr.Button("注册")
#         register_output = gr.Textbox(label="注册结果")
#         register_button.click(register, inputs=[name_input, email_input, password_input], outputs=register_output)
#
#
#     # 登录界面
#     with gr.Row():
#         login_email_input = gr.Textbox(label="邮箱", placeholder="请输入邮箱")
#         login_password_input = gr.Textbox(label="密码", placeholder="请输入密码", type="password")
#         login_button = gr.Button("登录")
#         login_output = gr.Textbox(label="登录结果")
#         login_button.click(
#             login,
#             inputs=[login_email_input, login_password_input],
#             outputs=login_output
#             )
#
#     # #模型选择
#     # with gr.Row():
#     #     model_dropdown = gr.Dropdown(
#     #         label="模型选择",
#     #         choices=model_labels,
#     #         value=model_labels[0],  # 默认选中第一个
#     #     )
#
#     # 内容生成界面（改造成聊天框）
#     chatbot = gr.Chatbot(label="对话历史")  # 聊天窗口
#     prompt_input = gr.Textbox(label="生成提示", placeholder="请输入要生成的内容提示")
#     generate_button = gr.Button("生成内容")
#
#     token = session_token
#     global model_labels
#     # 点击按钮时：更新对话框
#     generate_button.click(
#         generate_content,
#         inputs=[prompt_input, token_input, chatbot],  # token_input 是 gr.Textbox
#         outputs=chatbot
#     )
#     #成为会员
#     with gr.Row():
#         token = session_token
#         recharge_button = gr.Button("成为会员")
#         recharge_output = gr.Textbox(label="充值结果")
#         recharge_button.click(recharge, inputs=[token], outputs=recharge_output)
#
# demo.launch(share=True)
import gradio as gr
import requests


model_labels = ["deepseek-chat"]

# 注册函数
def register(name, email, password):
    url = "http://localhost:8080/register"
    data = {
        "name": name,
        "email": email,
        "password": password
    }
    response = requests.post(url, json=data)
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
    response = requests.post(url, json=data)
    if response.status_code == 200:
        token = response.json().get('token')
        return f"登录成功！Token: {token}"
    else:
        return f"登录失败: {response.json().get('error')}"

# 内容生成函数
def generate_content(prompt, history, token, model="deepseek-chat"):
    url = "http://localhost:8080/generate"
    headers = {
        "Authorization": f"Bearer {token}"
    }
    data = {
        "prompt": prompt,
        "model": model
    }
    response = requests.post(url, json=data, headers=headers)
    if response.status_code == 200:
        ai_response = response.json().get('result')
        history.append((prompt, ai_response))
        return history
    else:
        error = response.json().get('error', "生成失败")
        history.append((prompt, "❌ " + error))
        return history

def recharge(token):
    url = "http://localhost:8080/recharge"
    headers = {
        "Authorization": f"Bearer {token}"
    }
    response = requests.post(url, headers=headers)
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
    token_input = gr.Textbox(label="Token", placeholder="请输入登录后获得的 Token")
    generate_button = gr.Button("生成内容")

    # 点击按钮时：更新对话框
    generate_button.click(
        generate_content,
        inputs=[prompt_input, chatbot, token_input],
        outputs=chatbot
    )

    # 成为会员（充值）
    with gr.Row():
        token_input_recharge = gr.Textbox(label="Token", placeholder="请输入登录后获得的 Token")
        recharge_button = gr.Button("成为会员")
        recharge_output = gr.Textbox(label="充值结果")
        recharge_button.click(recharge, inputs=[token_input_recharge], outputs=recharge_output)

demo.launch(share=True)
