const std = @import("std");

var direct_paste: bool = false;
var scouter_name: []const u8 = "Zig";

fn mapHangingStatus(status: i32) []const u8 {
    return switch (status) {
        0 => "none",
        1 => "park",
        2 => "hang",
        else => "unknown",
    };
}

pub fn giveCorrection(json: []const u8) !TeamData {
    const parsed = try std.json.parseFromSlice(BadRoot, std.heap.page_allocator, json, .{});
    defer parsed.deinit();

    const badTeam: BadData = parsed.value.data;

    return .{
        .team = badTeam.Team,
        .match = .{
            .number = @intCast(badTeam.Match.Number),
            .isReplay = badTeam.Match.isReplay,
        },
        .scouter = scouter_name,
        .driverStation = .{
            .isBlue = badTeam.driverStation.IsBlue,
            .number = badTeam.driverStation.Number,
        },
        .auto = .{
            .canAuto = badTeam.Auto.Can,
            .hangAuto = badTeam.Auto.Hang,
            .scores = badTeam.Auto.Scores,
            .misses = badTeam.Auto.Misses,
            .ejects = badTeam.Auto.Ejects,
            .won = badTeam.Auto.@"Auto Won",
            .accuracy = .{
                .hpAccuracy = badTeam.Auto.@"Auto Human Player Accuracy",
                .robotAccuracy = badTeam.Auto.@"Auto Robot Accuracy",
            },
            .field = .{
                .left = badTeam.@"Auto Locations".@"Auto Field Left",
                .right = badTeam.@"Auto Locations".@"Auto Field Right",
                .mid = badTeam.@"Auto Locations".@"Auto Field Mid",
                .top = badTeam.@"Auto Locations".@"Auto Field Top",
                .bump = badTeam.@"Auto Locations".@"Auto Field Bump",
                .trench = badTeam.@"Auto Locations".@"Auto Field Trench",
                .didntCross = badTeam.@"Auto Locations".@"Auto Field DidntCross",
                .hp = badTeam.@"Auto Locations".@"Auto Field HP",
                .fuel = badTeam.@"Auto Locations".@"Auto Field Fuel",
            },
        },
        .teleop = .{
            .collection = .{
                .collectNeutral = badTeam.TeleOP.@"Neutral Collection",
                .collectHp = badTeam.TeleOP.@"Human Player Collection",
                .fuelCapacity = badTeam.TeleOP.@"Fuel Capacity",
            },
            .field = .{
                .bump = badTeam.@"TeleOp Locations".@"TeleOp Field Bump",
                .trench = badTeam.@"TeleOp Locations".@"TeleOp Field Trench",
            },
            .botType = badTeam.Misc.@"Bot Type",
            .playstyle = badTeam.Misc.@"Play Style",
        },
        .endgame = .{
            .park = mapHangingStatus(badTeam.Endgame.@"Hanging Status"),
            .climbTimer = badTeam.Endgame.Time,
            .endgameShoot = badTeam.Endgame.@"Shot During Endgame",
        },
        .issues = .{
            .disconnect = badTeam.Misc.@"Lost Communication Or Disabled",
            .loseTrack = badTeam.Misc.@"User Lost Track",
            .everBeached = badTeam.Misc.@"Ever Beached",
        },
        .notes = .{
            .perfNotes = badTeam.Notes.@"Performance Notes",
            .eventsNotes = badTeam.Notes.@"Event Notes",
            .commentsNotes = badTeam.Notes.Comments,
            .teleNotes = badTeam.Notes.@"TeleOp Notes",
            .autoNotes = badTeam.Notes.@"Auto Notes",
        },
        .rescouting = badTeam.Rescouting,
        .prescouting = false,
    };
}

pub fn outputJson(teamData: TeamData, stdout: ?*std.fs.File.Writer, output_path: ?[]const u8) !void {
    const formattedJson = std.json.fmt(teamData, .{ .whitespace = .indent_2 });
    if (stdout) |out| {
        try out.interface.print("{f}\n", .{formattedJson});
        try out.interface.flush();
    }

    if (output_path) |path_name| {
        var path_name_buffer: [std.fs.max_name_bytes]u8 = undefined;
        var arena = std.heap.FixedBufferAllocator.init(&path_name_buffer);
        const file_name = try std.fmt.allocPrint(arena.allocator(), "{s}{c}2026offseasons_{}_{s}{}_{}.json", .{ path_name, std.fs.path.sep, teamData.match.number, if (teamData.driverStation.isBlue) "blue" else "red", teamData.driverStation.number, std.time.milliTimestamp() });

        const file = try std.fs.cwd().createFile(file_name, .{});
        defer file.close();

        const written_json = try std.fmt.allocPrint(std.heap.page_allocator, "{f}\n", .{formattedJson});

        try file.writeAll(written_json);
    }
}

// used to find the json in batch when theyre squished up against one another
pub fn parseInputs(given_file: []const u8, stdout: ?*std.fs.File.Writer, output_path: ?[]const u8) !void {
    var found_next: bool = false;
    var bracket_count: usize = 0;
    var last_start: usize = 0;
    for (given_file, 0..) |char, i| {
        switch (char) { // quotes are too hard so imma assume they didnt use curly braces in comments because thats silly
            '{' => {
                bracket_count += 1;

                if (!found_next) {
                    found_next = true;
                    last_start = i;
                }
            },
            '}' => {
                bracket_count -= 1;
            },
            else => continue,
        }

        if (found_next and bracket_count == 0) {
            const json_target = given_file[last_start .. i + 1];
            var output_data: TeamData = undefined;

            std.debug.print("{s}\n", .{json_target});

            if (!direct_paste) {
                output_data = try giveCorrection(json_target);
                try outputJson(output_data, stdout, output_path);
            } else {
                const parsed = try std.json.parseFromSlice(TeamData, std.heap.page_allocator, json_target, .{ .ignore_unknown_fields = true, .duplicate_field_behavior = .use_first });
                defer parsed.deinit();
                output_data = parsed.value;
                output_data.scouter = scouter_name;
                try outputJson(output_data, stdout, output_path);
            }

            found_next = false;
        }
    }
}

pub fn main() !void {
    var stdout = std.fs.File.stdout().writerStreaming(&.{});
    defer stdout.interface.flush() catch |err| {
        std.debug.print("Failed to flush stdout: {}\n", .{err});
    };

    var argv = try std.process.argsWithAllocator(std.heap.page_allocator);
    defer argv.deinit();

    var inputJson: ?[]u8 = null;
    var output_path: ?[]const u8 = null;

    _ = argv.skip();
    while (argv.next()) |arg| {
        const save_flag = "--save";
        const name_flag = "--name";

        if (std.mem.startsWith(u8, arg, save_flag)) {
            if (arg.len > save_flag.len) { // --save=./path/to/save/dir  not file name just directory
                output_path = arg[save_flag.len + 1 ..];
            } else {
                output_path = ".";
            }
            continue;
        } else if (std.mem.startsWith(u8, arg, "--direct")) {
            direct_paste = true;
            continue;
        } else if (std.mem.startsWith(u8, arg, name_flag)) {
            if (arg.len > save_flag.len) { // --save=./path/to/save/dir  not file name just directory
                scouter_name = arg[save_flag.len + 1 ..];
            } else {
                scouter_name = "Zig";
            }
            continue;
        }

        inputJson = try std.fs.cwd().readFileAlloc(std.heap.page_allocator, arg, std.math.maxInt(usize));
    }

    if (inputJson == null) {
        inputJson = try std.fs.File.stdin().readToEndAlloc(std.heap.page_allocator, std.math.maxInt(usize));
    }

    try parseInputs(inputJson.?, &stdout, output_path);
}

pub const BadRoot = struct {
    key: []const u8,
    timestamp: i64,
    data: BadData,
};

pub const BadData = struct {
    Team: u64,
    Match: BadMatch,
    driverStation: BadDriverStation,
    Auto: BadAuto,
    @"Auto Locations": BadAutoLocations,
    Cycles: []BadCycle,
    TeleOP: BadTeleOP,
    Endgame: BadEndgame,
    @"TeleOp Locations": BadTeleOpLocations,
    Misc: BadMisc,
    Notes: BadNotes,
    Rescouting: bool,
};

pub const BadMatch = struct {
    Number: i32,
    isReplay: bool,
};

pub const BadDriverStation = struct {
    IsBlue: bool,
    Number: i32,
};

pub const BadAuto = struct {
    Can: bool,
    Hang: bool,
    Scores: i32,
    Misses: i32,
    Ejects: i32,
    @"Auto Human Player Accuracy": i32,
    @"Auto Robot Accuracy": i32,
    @"Auto Won": bool,
};

pub const BadAutoLocations = struct {
    @"Auto Field Left": bool,
    @"Auto Field Right": bool,
    @"Auto Field Mid": bool,
    @"Auto Field Top": bool,
    @"Auto Field Bump": bool,
    @"Auto Field Trench": bool,
    @"Auto Field DidntCross": bool,
    @"Auto Field HP": bool,
    @"Auto Field Fuel": bool,
};

pub const BadCycle = struct {
    Time: f64,
    Type: []const u8,
    activeHub: []const u8,
    Success: bool,
};

pub const BadTeleOP = struct {
    @"Neutral Collection": bool,
    @"Human Player Collection": bool,
    @"Fuel Capacity": []const u8,
};

pub const BadEndgame = struct {
    @"Hanging Status": i32,
    Time: f64,
    @"Shot During Endgame": bool,
};

pub const BadTeleOpLocations = struct {
    @"TeleOp Field Bump": bool,
    @"TeleOp Field Trench": bool,
};

pub const BadMisc = struct {
    @"Bot Type": []const u8,
    @"Play Style": []const u8,
    @"Lost Communication Or Disabled": bool,
    @"User Lost Track": bool,
    @"Ever Beached": bool,
};

pub const BadNotes = struct {
    @"Auto Notes": []const u8,
    @"TeleOp Notes": []const u8,
    @"Performance Notes": []const u8,
    @"Event Notes": []const u8,
    Comments: []const u8,
};

pub const TeamData = struct {
    team: u64,
    match: MatchInfo,
    scouter: []const u8 = "Zig",
    driverStation: DriverStationData,
    auto: AutoData,
    teleop: TeleopData,
    endgame: EndgameData,
    issues: IssuesData,
    notes: NotesData,
    rescouting: bool = false,
    prescouting: bool = false,
};

pub const AutoData = struct {
    canAuto: bool,
    hangAuto: bool,
    scores: i32,
    misses: i32,
    ejects: i32,
    won: bool,
    accuracy: AutoAccuracy,
    field: AutoField,
};

pub const AutoAccuracy = struct {
    hpAccuracy: i32,
    robotAccuracy: i32,
};

pub const AutoField = struct {
    left: bool,
    right: bool,
    mid: bool,
    top: bool,
    bump: bool,
    trench: bool,
    didntCross: bool,
    hp: bool,
    fuel: bool,
};

pub const TeleopData = struct {
    collection: CollectionData,
    field: TeleField,
    botType: []const u8,
    playstyle: []const u8,
};

pub const CollectionData = struct {
    collectNeutral: bool,
    collectHp: bool,
    fuelCapacity: []const u8,
};

pub const TeleField = struct {
    bump: bool,
    trench: bool,
};

pub const EndgameData = struct {
    park: []const u8,
    climbTimer: f64,
    endgameShoot: bool,
};

pub const IssuesData = struct {
    disconnect: bool,
    loseTrack: bool,
    everBeached: bool,
};

pub const NotesData = struct {
    perfNotes: []const u8,
    eventsNotes: []const u8,
    commentsNotes: []const u8,
    teleNotes: []const u8,
    autoNotes: []const u8,
};

pub const DriverStationData = struct {
    isBlue: bool,
    number: i32,
};

pub const MatchInfo = struct {
    number: u32,
    isReplay: bool,
};
