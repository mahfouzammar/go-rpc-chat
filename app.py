import streamlit as st

st.set_page_config(page_title="Go RPC Chat Monitor", page_icon="💬")

st.title("💬 غرفة المحادثة - Streamlit Interface")
st.write("مرحباً بك في واجهة التطبيق!")

# خانة لإدخال اسم المستخدم والرسالة
username = st.text_input("اسم المستخدم", value="User1")
message = st.text_area("اكتب رسالتك هنا...")

if st.button("إرسال الرسالة"):
    if message:
        st.success(f"تم إرسال الرسالة من {username}: {message}")
    else:
        st.warning("يرجى كتابة رسالة قبل الإرسال.")