import React, { useState } from "react";
import { Layout, Form, Input, Button, message, Tabs, Typography, Card } from "antd";
import axios from "axios";

const { Header, Footer, Content } = Layout;
const { Title, Text } = Typography;

const App = () => {
    const [token, setToken] = useState(null); // 保存用户登录后的 Token
    const [generatedText, setGeneratedText] = useState(""); // 保存生成的内容

    // 注册函数
    const handleRegister = async (values) => {
        try {
            const response = await axios.post("http://localhost:8080/register", {
                name: values.username,
                email: values.email,
                password: values.password,
            });
            message.success(`注册成功！用户ID: ${response.data.user_id}`);
        } catch (error) {
            message.error(error.response?.data?.error || "注册失败");
        }
    };

    // 登录函数
    const handleLogin = async (values) => {
        try {
            const response = await axios.post("http://localhost:8080/login", {
                email: values.email,
                password: values.password,
            });
            setToken(response.data.token); // 保存登录 Token
            message.success("登录成功！");
        } catch (error) {
            message.error(error.response?.data?.error || "登录失败");
        }
    };

    // 调用生成内容接口
    const handleGenerate = async (values) => {
        try {
            const response = await axios.post(
                "http://localhost:8080/generate",
                { prompt: values.prompt },
                { headers: { Authorization: `Bearer ${token}` } } // 添加 Token 到请求头
            );
            setGeneratedText(response.data.result); // 保存生成的内容
            message.success("内容生成成功！");
        } catch (error) {
            message.error(error.response?.data?.error || "生成失败");
        }
    };

    return (
        <Layout>
            <Header style={{ color: "white", fontSize: "20px" }}>
                用户注册、登录与内容生成
            </Header>
            <Content style={{ padding: "20px" }}>
                <Tabs defaultActiveKey="1">
                    {/* 注册页 */}
                    <Tabs.TabPane tab="注册" key="1">
                        <Card>
                            <Title level={3}>注册</Title>
                            <Form onFinish={handleRegister}>
                                <Form.Item
                                    label="用户名"
                                    name="username"
                                    rules={[{ required: true, message: "请输入用户名！" }]}
                                >
                                    <Input />
                                </Form.Item>
                                <Form.Item
                                    label="邮箱"
                                    name="email"
                                    rules={[{ required: true, message: "请输入邮箱！" }]}
                                >
                                    <Input />
                                </Form.Item>
                                <Form.Item
                                    label="密码"
                                    name="password"
                                    rules={[{ required: true, message: "请输入密码！" }]}
                                >
                                    <Input.Password />
                                </Form.Item>
                                <Form.Item>
                                    <Button type="primary" htmlType="submit">
                                        注册
                                    </Button>
                                </Form.Item>
                            </Form>
                        </Card>
                    </Tabs.TabPane>

                    {/* 登录页 */}
                    <Tabs.TabPane tab="登录" key="2">
                        <Card>
                            <Title level={3}>登录</Title>
                            <Form onFinish={handleLogin}>
                                <Form.Item
                                    label="邮箱"
                                    name="email"
                                    rules={[{ required: true, message: "请输入邮箱！" }]}
                                >
                                    <Input />
                                </Form.Item>
                                <Form.Item
                                    label="密码"
                                    name="password"
                                    rules={[{ required: true, message: "请输入密码！" }]}
                                >
                                    <Input.Password />
                                </Form.Item>
                                <Form.Item>
                                    <Button type="primary" htmlType="submit">
                                        登录
                                    </Button>
                                </Form.Item>
                            </Form>
                        </Card>
                    </Tabs.TabPane>

                    {/* 内容生成页 */}
                    <Tabs.TabPane tab="生成内容" key="3" disabled={!token}>
                        <Card>
                            <Title level={3}>生成内容</Title>
                            {!token && (
                                <Text type="danger">
                                    您需要先登录后才能使用此功能！
                                </Text>
                            )}
                            <Form onFinish={handleGenerate}>
                                <Form.Item
                                    label="生成提示"
                                    name="prompt"
                                    rules={[{ required: true, message: "请输入生成提示！" }]}
                                >
                                    <Input />
                                </Form.Item>
                                <Form.Item>
                                    <Button type="primary" htmlType="submit" disabled={!token}>
                                        生成
                                    </Button>
                                </Form.Item>
                            </Form>
                            {generatedText && (
                                <Card>
                                    <Title level={4}>生成结果：</Title>
                                    <Text>{generatedText}</Text>
                                </Card>
                            )}
                        </Card>
                    </Tabs.TabPane>
                </Tabs>
            </Content>
            <Footer style={{ textAlign: "center" }}>
                用户系统 ©2025 Created by 您的代码助手
            </Footer>
        </Layout>
    );
};

export default App;