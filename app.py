import json
import urllib.request
import streamlit as st

# عنوان سيرفر Go RPC (تأكد من تعديل الـ IP أو البورت حسب مكان استضافة سيرفر Go)
GO_SERVER_URL = "http://localhost:1234/_goRPC_"

st.set_page_config(page_title="Go RPC Chat Room", page_icon="💬", layout="centered")


# دالة مساعدة لإرسال طلبات RPC إلى سيرفر Go
def call_go_rpc(method, params):
    payload = json.dumps({"method": method, "params": [params], "id": 1}).encode(
        "utf-8"
    )
    req = urllib.request.Request(
        GO_SERVER_URL,
        data=payload,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=3) as response:
            res = json.loads(response.read().decode("utf-8"))
            return res.get("result")
    except Exception as e:
        return None


# دالة لجلب الرسائل من سيرفر Go
def fetch_messages_from_go():
    # استدعاء ChatService.GetMessages
    res = call_go_rpc("ChatService.GetMessages", {})
    if res is not None:
        return res
    return []


# إدارة جلسة اسم المستخدم
if "username" not in st.session_state:
    st.session_state.username = None

# --- الشاشة الأولى: تسجيل الدخول ---
if not st.session_state.username:
    st.title("💬 مرحباً بك في غرفة المحادثة")
    st.subheader("يرجى إدخال اسمك للانضمام إلى المحادثة")

    input_name = st.text_input(
        "اسم المستخدم (Username):", placeholder="اكتب اسمك هنا..."
    )

    if st.button("انضمام للدردشة 🚀"):
        if input_name.strip():
            st.session_state.username = input_name.strip()
            # إرسال إشعار انضمام إلى سيرفر Go
            join_msg = {
                "Sender": "System",
                "Text": f"*** {st.session_state.username} انضم إلى غرفة المحادثة ***",
            }
            call_go_rpc("ChatService.PostMessage", join_msg)
            st.rerun()
        else:
            st.warning("رجاءً اكتب اسم مستخدم صحيح قبل الاستمرار.")

# --- الشاشة الثانية: واجهة المحادثة الجماعية ---
else:
    st.title("💬 غرفة المحادثة الجماعية (Go RPC)")
    st.caption(f"متصل حالياً باسم: **{st.session_state.username}**")

    # زر تحديث لجلب أحدث الرسائل المتبادلة في أي وقت
    if st.button("🔄 تحديث الرسائل"):
        st.rerun()

    st.divider()

    # جلب جميع الرسائل الموحدة المخزنة في سيرفر Go
    messages = fetch_messages_from_go()

    for msg in messages:
        sender = msg.get("Sender", "")
        text = msg.get("Text", "")

        if sender == "System":
            st.info(text)
        else:
            is_me = sender == st.session_state.username
            role_type = "user" if is_me else "assistant"

            with st.chat_message(role_type):
                st.write(f"**{sender}**: {text}")

    # صندوق إرسال الرسائل
    if prompt := st.chat_input("اكتب رسالتك هنا..."):
        new_msg = {"Sender": st.session_state.username, "Text": prompt}
        # إرسال الرسالة إلى سيرفر Go المركزي
        call_go_rpc("ChatService.PostMessage", new_msg)
        st.rerun()