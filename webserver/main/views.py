# coding=utf-8
from . import admin
from flask import render_template, request, redirect, url_for
from model import DBSession, Task, TaskInstance, User
from flask.ext.login import login_user, login_required, logout_user, current_user
from webserver import login_manager
from datetime import datetime
import json


@login_manager.user_loader
def load_user(user_id):
    session = DBSession()
    uid = session.query(User).get(int(user_id))
    session.close()
    return uid


@login_manager.unauthorized_handler
def unauthorized():
    return redirect('/login')


@admin.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'GET':
        return render_template('login.html')
    else:
        user_name = request.form.get('user_name')
        password = request.form.get('password')
        session = DBSession()
        user = session.query(User).filter_by(name=user_name).first()
        session.close()
        if user is not None and user.verify_password(password):
            login_user(user)
        return redirect('/')


@admin.route('/logout')
@login_required
def logout():
    logout_user()
    return redirect('/login')


@admin.route('/')
@login_required
def home():
    session = DBSession()
    tasks = session.query(Task).order_by(Task.id.desc()).all()
    session.close()
    return render_template('list_task.html', tasks=tasks)


@admin.route('/new_task', methods=['POST', 'GET'])
@login_required
def new_task():
    if request.method == 'GET':
        extend_id = request.args.get('extend_id')
        return render_template('new_task.html')
    else:
        data = request.form
        task = Task(
                name=data.get('task_name'),
                user=current_user.name,
                create_time=datetime.now(),
                command=data.get('command'),
                priority=data.get('priority'),
                machine_pool=data.get('machine_pool').split('\n'),
                father_task=data.get('father_task').split('\n'),
                valid=data.get('valid') == 'true',
                rerun=data.get('rerun') == 'true',
                rerun_times=data.get('rerun_times'),
                scheduled_type=data.get('scheduled_type'),
                year=data.get('year'),
                month=data.get('month'),
                day=data.get('day'),
                hour=data.get('hour'),
                minute=data.get('minute')
        )
        session = DBSession()
        session.add(task)
        session.commit()
        session.close()
        return '/' if request.form.get('has_next') == 'false' else 'new_task'


@admin.route('/list_execute_log')
@login_required
def list_execute_log():
    session = DBSession()
    tasks_instance = session.query(TaskInstance, Task) \
        .join(Task, TaskInstance.task_id == Task.id) \
        .order_by(TaskInstance.id.desc()).all()
    session.close()
    return render_template('list_execute_log.html', tasks_instance=tasks_instance)


@admin.route('/list_action_log')
@login_required
def list_action_log():
    return render_template('list_action_log.html')


@admin.route('/list_user')
@login_required
def list_user():
    session = DBSession()
    users = session.query(User).order_by(User.id.desc()).all()
    session.close()
    return render_template('list_user.html', users=users)


@admin.route('/new_user', methods=['GET', 'POST'])
@login_required
def new_user():
    if request.method == 'GET':
        return render_template('new_user.html')
    else:
        user = User()
        user.name = request.form.get('user_name')
        user.password = request.form.get('password')
        user.email = request.form.get('email')
        session = DBSession()
        session.add(user)
        session.commit()
        session.close()
        return redirect('/list_user')


@admin.route('/list_machine')
@login_required
def list_machine():
    return render_template('list_machine.html')


@admin.route('/list_machine_status')
@login_required
def list_machine_status():
    return render_template('list_machine_status.html')


@admin.route('/new_machine')
@login_required
def new_machine():
    return render_template('new_machine.html')
