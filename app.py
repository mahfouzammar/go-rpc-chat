import streamlit as st

st.set_page_config(page_title="Go RPC Chat Room", page_icon="💬", layout="centered")

# استخدام st.session_state بشكل مركزي بدلاً من الـ cache المشترك لتجنب انفصال الذاكرة بين الأجهزة على السحابة
if "global_chat" not in st.session_state:
    st.session_state.global_chat = []

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
            st.session_state.global_chat.append(join_msg)
            st.rerun()
        else:
            st.warning("رجاءً اكتب اسم مستخدم صحيح قبل الاستمرار.")

# --- الشاشة الثانية: واجهة المحادثة الجماعية ---
else:
    st.title("💬 غرفة المحادثة الجماعية")
    st.caption(f"متصل حالياً باسم: **{st.session_state.username}**")
    
    st.divider()

    # عرض جميع الرسائل الموجودة في الجلسة الحالية
    for msg in st.session_state.global_chat:
        if msg["role"] == "system":
            st.info(msg["text"])
        else:
            is_me = msg["sender"] == st.session_state.username
            role_type = "user" if is_me else "assistant"
            
            with st.chat_message(role_type):
                st.write(f"**{msg['sender']}**: {msg['text']}")

    # صندوق إرسال الرسائل
    if prompt := st.chat_input("اكتب رسالتك هنا..."):
        new_msg = {
            "role": "user",
            "sender": st.session_state.username,
            "text": prompt
        }
        st.session_state.global_chat.append(new_msg)
        st.rerun()