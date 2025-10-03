import gradio as gr
import requests

session = requests.Session()
model_labels = ["deepseek-chat"]
print(session.cookies)

# 获取用户所有会话
def get_conversations():
    url = "http://localhost:8080/getConversation"
    response = session.get(url)
    if response.status_code == 200:
        conversations = response.json().get('conversations')
        #return [(title,id)]
        return {c['title']: c['conversation_id'] for c in conversations}
    else:
        return {}
    
#新建会话
def new_conversation():
    url = "http://localhost:8080/newConversation"
    response = session.post(url)
    if response.status_code == 200:
        data = response.json()
        return f"成功新建新会话: {data.get('conversation_id')}-{data.get('title')}"
    else:
        return {}

def load_conversation_history(conversation_id):
    url = f"http://localhost:8080/history/{conversation_id}"
    response = session.get(url)
    if response.status_code == 200:
        history = response.json().get('history', [])
        chat = []
        current_user_msg = None

        for h in history:
            if h['role'] == 'user':
                # 遇到用户消息，暂存
                current_user_msg = h['message']
            elif h['role'] == 'ai':
                # 遇到 AI 消息，把它和最近的用户消息配对
                if current_user_msg is not None:
                    chat.append((current_user_msg, h['message']))
                    current_user_msg = None
                else:
                    # 如果没有用户消息，就单独显示 AI 回复
                    chat.append((None, h['message']))

        # 防止最后一个用户消息没有配对 AI 回复
        if current_user_msg is not None:
            chat.append((current_user_msg, None))

        return chat
    else:
        return [("获取该会话历史记录失败", None)]

# #显示历史记录
# def get_history():
#     url = "http://localhost:8080/history"
#     response = session.get(url)
#     if response.status_code == 200:
#         datas = response.json().get('history')
#         history = []
#         user_chat = None
#         for data in datas:
#             if data['role'] == 'user':
#                 user_chat = data['message']
#             elif data['role'] == 'ai' and user_chat is not None:
#                 history.append((user_chat, data['message']))
#                 user_chat = None
#         return history
#     else:
#         return "获取历史记录失败"


#模型选择
def get_models():
    url = "http://localhost:8080/getModel"
    response = session.get(url)
    if response.status_code == 200:
        print(response.json())
        available_models = response.json().get('availableModel', ["deepseek-chat"])
        print(available_models)
        return gr.Dropdown(choices=available_models, value=available_models[0])
    else:
        return gr.Dropdown(choices=["deepseek-chat"], value="deepseek-chat")


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

# 退出登录
def logout():
    url = "http://localhost:8080/logout"
    response = session.post(url)
    if response.status_code == 200:
        return "登出成功"
    else:
        return "登出失败"
# 内容生成函数
def generate_content(prompt, history, model, conversation_id):
    url = "http://localhost:8080/generate"
    data = {
        "prompt": prompt,
        "model": model,
        "conversation_id": conversation_id
    }
    response = session.post(url, json=data)
    if response.status_code == 200:
        ai_response = response.json().get('result')
        history.append((prompt, ai_response))
    else:
        error = response.json().get('error', "生成失败")
        history.append((prompt, "❌ " + error))
    return history

# 充值函数
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

    with gr.Row():
        #左边: 会话管理区
        with gr.Column(scale=1):
            gr.Markdown("##会话管理")
            conv_list = gr.Dropdown(label="选择会话", choices=[], interactive=True)
            refresh_btn = gr.Button("刷新会话")
            new_conv_btn = gr.Button("新建会话")
            conv_status = gr.Textbox(label="会话状态", interactive=False)
        #右边: 聊天区
        with gr.Column(scale=3):
            gr.Markdown("##聊天窗口")
            chatbot = gr.Chatbot(label="聊天记录", height=500)
            prompt_input = gr.Textbox(label="输入信息", placeholder="输入内容后按下回车")

    conv_list.change(
        fn=load_conversation_history,
        inputs=conv_list,     # 传入选择的会话 ID
        outputs=chatbot       # 更新聊天框
    )

    # 刷新会话列表
    def refresh_conv():
        convs = get_conversations()
        if convs:
            return gr.update(choices=list(convs.values()), label="选择会话", value=list(convs.values())[0]), "✅ 获取会话成功"
        else:
            return gr.update(choices=[], value=None), "❌ 没有会话"
    #刷新对话
    refresh_btn.click(refresh_conv, outputs=[conv_list, conv_status])
    #新建会话
    new_conv_btn.click(new_conversation, outputs=conv_status).then(refresh_conv, outputs=[conv_list, conv_status])

    # # 内容生成界面（改造成聊天框）
    # chatbot = gr.Chatbot(label="对话历史")  # 聊天窗口
    # #查看历史记录
    # history_button = gr.Button("查看历史记录")
    # history_button.click(fn=get_history, outputs=chatbot)
    # prompt_input = gr.Textbox(label="生成提示", placeholder="请输入要生成的内容提示")

    #模型选择
    model_dropdown = gr.Dropdown(label="选择模型", choices=[])
    refresh_models_button = gr.Button("刷新可用模型")
    refresh_models_button.click(fn=get_models, outputs=model_dropdown)
    generate_button = gr.Button("生成内容")

    # 点击按钮时：更新对话框
    print(model_dropdown)
    generate_button.click(
        generate_content,
        inputs=[prompt_input, chatbot, model_dropdown, conv_list],
        outputs=chatbot
    )

    # 成为会员（充值）
    with gr.Row():
        recharge_button = gr.Button("成为会员")
        recharge_output = gr.Textbox(label="充值结果")
        recharge_button.click(recharge, outputs=recharge_output)

    # 登出按钮
    with gr.Row():
        logout_button = gr.Button("退出登录")
        logout_output = gr.Textbox(label="登出结果")
        logout_button.click(logout, outputs=logout_output)

demo.launch(share=True)
