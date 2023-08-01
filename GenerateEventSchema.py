import os

d = 'bin/events'
out = 'begin;insert into event_quests(max_players, quest_type, quest_id, mark)values'
for y, _, x in os.walk(d):
    for z in x:
        f = os.path.join(y, z)
        if os.stat(f).st_size < 352:
            continue
        with open(f, 'rb') as p:
            p.seek(9, 0)
            max_players = int.from_bytes(p.read(1), 'big')
            quest_type = int.from_bytes(p.read(1), 'big')
            p.seek(14, 0)
            mark = int.from_bytes(p.read(4), 'big')
            p.seek(68, 0)
            quest_id = int.from_bytes(p.read(2), 'big')
            out += f'({max_players},{quest_type},{quest_id},{mark}),'
with open('PortedEventQuests.sql', 'w') as f:
    f.write(out[:-1]+';end')
