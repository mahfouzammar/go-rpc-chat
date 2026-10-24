import streamlit as st
import time

st.set_page_config(page_title="Go RPC Chat Room", page_icon="💬", layout="centered")

# 1. مخزن الرسائل الجماعي المشترك
@st.cache_resource
def get_global_chat_history():
    return []

global_messages = get_global_chat_history()

# 2. إدارة جلسة المستخدم
if "username" not in st.session_state:
    st.session_state.username = None

# --- الشاشة الأولى: تسجيل الدخول ---
if not st.session_state.username:
    st.title("💬 مرحباً بك في غرفة المحادثة")
    st.subheader("يرجى إدخال اسمك للانضمام إلى المحادثة")
    
    input_name = st.text_input("اسم المستخدم (Username):", placeholder="اكتب اسمك هنا...")
    
    if st.button("انضمام للدردشة 🚀"):
        if input_name.strip():
            st.session_state.username = input_name.strip()
            join_msg = {
                "role": "system",
                "sender": "System",
                "text": f"*** {st.session_state.username} انضم إلى غرفة المحادثة ***"
            }
            global_messages.append(join_msg)
            st.rerun()
        else:
            st.warning("رجاءً اكتب اسم مستخدم صحيح قبل الاستمرار.")

# --- الشاشة الثانية: واجهة المحادثة الجماعية ---
else:
    st.title("💬 غرفة المحادثة الجماعية")
    st.caption(f"متصل حالياً باسم: **{st.session_state.username}**")
    
    st.divider()

    # عرض جميع الرسائل مباشرة
    for msg in global_messages:
        if msg["role"] == "system":
            st.info(msg["text"])
        else:
            is_me = msg["sender"] == st.session_state.username
            role_type = "user" if is_me else "assistant"
            
            with st.chat_message(role_type):
                st.write(f"**{msg['sender']}**: {msg['text']}")

    # صندوق إرسال الرسائل الجديد
    if prompt := st.chat_input("اكتب رسالتك هنا..."):
        new_msg = {
            "role": "user",
            "sender": st.session_state.username,
            "text": prompt
        }
        global_messages.append(new_msg)
        # إعادة تحميل الصفحة فوراً لكي تظهر رسالتك في الشات حالاً دون انتظار أحد
        st.rerun()